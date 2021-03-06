/*
Copyright 2020 The Magma Authors.

This source code is licensed under the BSD-style license found in the
LICENSE file in the root directory of this source tree.

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mock_pgw

import (
	"fmt"
	"math/rand"
	"net"
	"strings"

	"github.com/pkg/errors"
	"github.com/wmnsk/go-gtp/gtpv2"
	"github.com/wmnsk/go-gtp/gtpv2/ie"
	"github.com/wmnsk/go-gtp/gtpv2/message"
)

func getHandleCreateSessionRequest() gtpv2.HandlerFunc {
	return func(c *gtpv2.Conn, sgwAddr net.Addr, msg message.Message) error {
		fmt.Println("mock PGW Received a CreateSessionRequest")

		session := gtpv2.NewSession(sgwAddr, &gtpv2.Subscriber{Location: &gtpv2.Location{}})
		bearer := session.GetDefaultBearer()

		var err error
		csReqFromSGW := msg.(*message.CreateSessionRequest)
		if imsiIE := csReqFromSGW.IMSI; imsiIE != nil {
			imsi, err := imsiIE.IMSI()
			if err != nil {
				return err
			}
			session.IMSI = imsi

			// remove previous session for the same subscriber if exists.
			sess, err := c.GetSessionByIMSI(imsi)
			if err != nil {
				switch err.(type) {
				case *gtpv2.UnknownIMSIError:
					// whole new session. just ignore.
				default:
					return errors.Wrap(err, "got something unexpected")
				}
			} else {
				c.RemoveSession(sess)
			}
		} else {
			return &gtpv2.RequiredIEMissingError{Type: ie.IMSI}
		}
		if msisdnIE := csReqFromSGW.MSISDN; msisdnIE != nil {
			session.MSISDN, err = msisdnIE.MSISDN()
			if err != nil {
				return err
			}
		} else {
			return &gtpv2.RequiredIEMissingError{Type: ie.MSISDN}
		}
		if meiIE := csReqFromSGW.MEI; meiIE != nil {
			session.IMEI, err = meiIE.MobileEquipmentIdentity()
			if err != nil {
				return err
			}
		} else {
			return &gtpv2.RequiredIEMissingError{Type: ie.MobileEquipmentIdentity}
		}
		if apnIE := csReqFromSGW.APN; apnIE != nil {
			bearer.APN, err = apnIE.AccessPointName()
			if err != nil {
				return err
			}
		} else {
			return &gtpv2.RequiredIEMissingError{Type: ie.AccessPointName}
		}
		if netIE := csReqFromSGW.ServingNetwork; netIE != nil {
			session.MNC, err = netIE.MNC()
			if err != nil {
				return err
			}
		} else {
			return &gtpv2.RequiredIEMissingError{Type: ie.ServingNetwork}
		}
		if ratIE := csReqFromSGW.RATType; ratIE != nil {
			session.RATType, err = ratIE.RATType()
			if err != nil {
				return err
			}
		} else {
			return &gtpv2.RequiredIEMissingError{Type: ie.RATType}
		}
		if sgwTEID := csReqFromSGW.SenderFTEIDC; sgwTEID != nil {
			teid, err := sgwTEID.TEID()
			if err != nil {
				return err
			}
			session.AddTEID(gtpv2.IFTypeS5S8SGWGTPC, teid)
		} else {
			return &gtpv2.RequiredIEMissingError{Type: ie.FullyQualifiedTEID}
		}

		if brCtxIE := csReqFromSGW.BearerContextsToBeCreated; brCtxIE != nil {
			for _, childIE := range brCtxIE.ChildIEs {
				switch childIE.Type {
				case ie.EPSBearerID:
					bearer.EBI, err = childIE.EPSBearerID()
					if err != nil {
						return err
					}
				case ie.FullyQualifiedTEID:
					it, err := childIE.InterfaceType()
					if err != nil {
						return err
					}
					// only used for user plane
					teidOut, err := childIE.TEID()
					if err != nil {
						return err
					}
					session.AddTEID(it, teidOut)
				}
			}
		} else {
			return &gtpv2.RequiredIEMissingError{Type: ie.BearerContext}
		}

		bearer.SubscriberIP = getRandomIp()

		// create PGW side C and U TEIDs
		cIP := strings.Split(c.LocalAddr().String(), ":")[0]
		pgwFTEIDc := c.NewSenderFTEID(cIP, "").WithInstance(1)
		uIP := strings.Split(dummyUserPlanePgwIP, ":")[0]
		pgwFTEIDu := ie.NewFullyQualifiedTEID(gtpv2.IFTypeS5S8PGWGTPU, rand.Uint32(), uIP, "")

		sgwTEIDc, err := session.GetTEID(gtpv2.IFTypeS5S8SGWGTPC)
		// send
		csRspFromPGW := message.NewCreateSessionResponse(
			sgwTEIDc, 0,
			ie.NewCause(gtpv2.CauseRequestAccepted, 0, 0, 0, nil),
			pgwFTEIDc,
			ie.NewPDNAddressAllocation(bearer.SubscriberIP),
			ie.NewAPNRestriction(gtpv2.APNRestrictionPublic2),
			ie.NewBearerContext(
				ie.NewCause(gtpv2.CauseRequestAccepted, 0, 0, 0, nil),
				ie.NewEPSBearerID(bearer.EBI),
				pgwFTEIDu,
				ie.NewChargingID(bearer.ChargingID)))

		session.AddTEID(gtpv2.IFTypeS5S8PGWGTPC, pgwFTEIDc.MustTEID())
		session.AddTEID(gtpv2.IFTypeS5S8PGWGTPU, pgwFTEIDu.MustTEID())

		if err := c.RespondTo(sgwAddr, csReqFromSGW, csRspFromPGW); err != nil {
			return err
		}
		s5pgwTEID, err := session.GetTEID(gtpv2.IFTypeS5S8PGWGTPC)
		if err != nil {
			return err
		}
		c.RegisterSession(s5pgwTEID, session)
		if err := session.Activate(); err != nil {
			return err
		}
		if err := session.Activate(); err != nil {
			return err
		}

		return nil
	}
}

func getRandomIp() string {
	return fmt.Sprintf("192.168.1.%d", (1 + rand.Intn(250)))
}
